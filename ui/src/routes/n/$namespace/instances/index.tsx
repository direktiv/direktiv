import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/instances/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/instances/"!</div>
}
