import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/explorer/workflow/edit/$')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/n/$namespace/explorer/workflow/edit/$filename"!</div>
}
