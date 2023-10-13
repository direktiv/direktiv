import Notification from "~/design/Notification";
import { apiFactory } from "~/api/apiFactory";
import { useNamespace } from "~/util/store/namespace";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { z } from "zod";

interface NotificationMenuProps {
  className?: string;
  text?: string;
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

// const getText

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

const NotificationMenu: React.FC<NotificationMenuProps> = () => {
  const namespace = useNamespace();
  const { t } = useTranslation();
  const text = t("components.notificationMenu.hasIssues.secrets.text");
  //  const text = t("components.userMenu.loggedIn");
  const { data, isLoading } = useQuery({
    queryKey: ["lint", namespace],
    queryFn: () =>
      getNamespaceLinting({
        urlParams: { namespace: namespace ?? "" },
      }),
  });

  const showIndicator = !!data?.issues.length;

  return <Notification showIndicator={true} isLoading={false} text={text} />;
};

export default NotificationMenu;
