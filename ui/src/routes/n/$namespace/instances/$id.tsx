import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/instances/$id')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/instances/$id"!</div>
}
