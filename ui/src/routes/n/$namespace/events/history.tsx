import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/events/history')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/events/history"!</div>
}
