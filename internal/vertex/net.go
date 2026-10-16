package vertex

import "AnonymousVertex/internal/transport"

func (c *VertexAIClient) Net() *transport.NetworkClient { return c.net }
