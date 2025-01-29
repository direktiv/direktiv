import { createFileRoute } from '@tanstack/react-router'

const ServicesPage = () => (
  <div>
    <h1>Services page</h1>
  </div>
)

export const Route = createFileRoute('/n/$namespace/services/')({
  component: ServicesPage,
})
