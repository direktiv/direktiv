import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/mirror/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/mirror/"!</div>
}
