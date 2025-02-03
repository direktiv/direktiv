import EndpointPage from '~/pages/namespace/Explorer/Endpoint'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/explorer/endpoint/$')({
  component: EndpointPage,
})
