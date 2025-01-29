import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/gateway/consumers')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/gateway/consumers"!</div>
}
