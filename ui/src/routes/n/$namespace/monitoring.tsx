import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/monitoring')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/monitoring"!</div>
}
