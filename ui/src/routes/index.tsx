import OnboardingPage from "~/pages/OnboardingPage";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: OnboardingPage,
});
