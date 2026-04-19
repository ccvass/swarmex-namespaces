package namespaces

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

const (
	labelNamespace = "swarmex.namespace"
	labelIsolated  = "swarmex.namespace.isolated"
	networkPrefix  = "ns-"
)

type Controller struct {
	client   *client.Client
	logger   *slog.Logger
	networks map[string]bool
	pending  map[string]bool
	mu       sync.Mutex
}

func New(cli *client.Client, logger *slog.Logger) *Controller {
	return &Controller{client: cli, logger: logger, networks: make(map[string]bool), pending: make(map[string]bool)}
}

func (c *Controller) HandleEvent(ctx context.Context, event events.Message) {
	if event.Type != events.ServiceEventType {
		return
	}
	if event.Action != events.ActionCreate && event.Action != events.ActionUpdate {
		return
	}
	// Debounce: skip if already processing this service
	c.mu.Lock()
	if c.pending[event.Actor.ID] {
		c.mu.Unlock()
		return
	}
	c.pending[event.Actor.ID] = true
	c.mu.Unlock()

	go func() {
		time.Sleep(2 * time.Second) // debounce
		c.reconcile(ctx, event.Actor.ID)
		c.mu.Lock()
		delete(c.pending, event.Actor.ID)
		c.mu.Unlock()
	}()
}

func (c *Controller) reconcile(ctx context.Context, serviceID string) {
	svc, _, err := c.client.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{})
	if err != nil {
		return
	}
	ns, ok := svc.Spec.Labels[labelNamespace]
	if !ok || ns == "" {
		return
	}

	netName := networkPrefix + ns
	c.ensureNetwork(ctx, netName)

	// Connect service to namespace network
	for _, n := range svc.Spec.TaskTemplate.Networks {
		if n.Target == netName {
			return // already connected
		}
	}

	svc.Spec.TaskTemplate.Networks = append(svc.Spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: netName})

	// If isolated, remove all other non-system networks
	if svc.Spec.Labels[labelIsolated] == "true" {
		var filtered []swarm.NetworkAttachmentConfig
		for _, n := range svc.Spec.TaskTemplate.Networks {
			if n.Target == netName || n.Target == "ingress" {
				filtered = append(filtered, n)
			}
		}
		svc.Spec.TaskTemplate.Networks = filtered
	}

	_, err = c.client.ServiceUpdate(ctx, serviceID, svc.Version, svc.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		c.logger.Error("namespace update failed", "service", svc.Spec.Name, "namespace", ns, "error", err)
		return
	}
	c.logger.Info("service assigned to namespace", "service", svc.Spec.Name, "namespace", ns, "isolated", svc.Spec.Labels[labelIsolated])
}

func (c *Controller) ensureNetwork(ctx context.Context, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.networks[name] {
		return
	}
	_, err := c.client.NetworkCreate(ctx, name, network.CreateOptions{
		Driver:     "overlay",
		Attachable: true,
		Labels:     map[string]string{"swarmex.managed": "true"},
	})
	if err != nil {
		// might already exist
		c.networks[name] = true
		return
	}
	c.networks[name] = true
	c.logger.Info("namespace network created", "network", name)
}
