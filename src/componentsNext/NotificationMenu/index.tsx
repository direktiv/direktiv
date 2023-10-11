import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Notification from "~/design/Notification";
import { apiFactory } from "~/api/apiFactory";
import { twMergeClsx } from "~/util/helpers";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { z } from "zod";

interface NotificationMenuProps {
  className?: string;
}

/**
 

example
{
  "namespace":  {
    "createdAt":  "2023-10-04T07:43:17.082556Z",
    "updatedAt":  "2023-10-04T07:43:17.082556Z",
    "name":  "dir-672",
    "oid":  "0cd5c136-5e53-40cc-aa46-0f68a9d5983c",
    "notes":  {
      "commit_hash":  "3c6d83c6e852c1197e84e2fa474d7f70f46065e3",
      "ref":  "main",
      "url":  "https://github.com/direktiv/direktiv-examples.git"
    }
  },
  "issues":  [
    {
      "type":  "secret",
      "id":  "ACCESS_KEY",
      "issue":  "secret 'ACCESS_KEY' has not been initialized",
      "level":  "critical"
    },
  
  ]
}

 */

const IssueSchema = z.object({
  type: z.enum(["secret"]),
  id: z.string(),
  issue: z.string(),
  level: z.string(),
});

const LintSchema = z.object({
  issues: z.array(IssueSchema),
});

export const getNamespaceLinting = apiFactory({
  url: ({ namespace }: { namespace: string }) =>
    `/api/namespaces/${namespace}/lint`,
  method: "GET",
  schema: LintSchema,
});

const NotificationMenu: React.FC<NotificationMenuProps> = ({ className }) => {
  const namespace = useNamespace();
  const { data, isLoading } = useQuery({
    queryKey: ["lint", namespace],
    queryFn: () =>
      getNamespaceLinting({
        urlParams: { namespace: namespace ?? "" },
      }),
  });

  const showIndicator = !!data?.issues.length;

  return (
    <div className={twMergeClsx("self-end text-right", className)}>
      <Popover>
        <PopoverTrigger>
          <Notification hasMessage={showIndicator} />
        </PopoverTrigger>
        <PopoverContent align="end" className="p-4">
          {isLoading && "loading"}
          Place content for the popover here.
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default NotificationMenu;
