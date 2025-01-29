import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/mirror/logs/$id')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/mirror/logs/$id"!</div>
}
