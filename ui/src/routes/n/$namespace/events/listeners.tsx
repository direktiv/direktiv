import ListenersList from '~/pages/namespace/Events/Listeners'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/n/$namespace/events/listeners')({
  component: ListenersList,
})
