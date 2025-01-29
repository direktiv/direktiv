import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/explorer/tree/$path')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/explorer/tree/$path"!</div>
}
