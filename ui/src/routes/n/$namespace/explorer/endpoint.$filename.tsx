import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute(
  '/n/$namespace/explorer/endpoint/$filename',
)({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/explorer/endpoint/$filename"!</div>
}
