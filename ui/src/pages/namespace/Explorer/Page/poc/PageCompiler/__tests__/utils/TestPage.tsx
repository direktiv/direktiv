import { DirektivPagesType } from "../../../schema";
import { LiveLayout } from "../../../PageLayout/LiveLayout";
import { QueryClientProvider } from "@tanstack/react-query";
import queryClient from "~/util/queryClient";

export const TestLivePage = ({ page }: { page: DirektivPagesType }) => (
  <QueryClientProvider client={queryClient}>
    <LiveLayout page={page} />
  </QueryClientProvider>
);
