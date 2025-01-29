import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/settings')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/settings"!</div>
}
