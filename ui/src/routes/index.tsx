import { createFileRoute } from "@tanstack/react-router";

const HomePage = () => (
  <div>
    <h1>Welcome to the Home Page!</h1>
    <p>
      This is the home page of your app, which is routed correctly with TanStack
      Router.
    </p>
  </div>
);

export const Route = createFileRoute("/")({
  component: HomePage,
});
