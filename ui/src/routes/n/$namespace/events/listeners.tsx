import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/events/listeners')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/events/listeners"!</div>
}
