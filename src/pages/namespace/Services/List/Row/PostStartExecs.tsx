import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import Badge from "~/design/Badge";
import { FC } from "react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import { useTranslation } from "react-i18next";

export type PostStartExecsProps = {
  exec: ServiceSchemaType["post_start_exec"];
};

const PostStartExecs: FC<PostStartExecsProps> = ({ exec }) => {
  const { t } = useTranslation();
  const execCount = exec?.length ?? 0;

  return execCount > 0 ? (
    <HoverCard>
      <HoverCardTrigger className="inline-flex">
        <Badge variant="secondary">
          {t("pages.services.list.tableRow.execLabel")}
        </Badge>
      </HoverCardTrigger>
      <HoverCardContent className="flex flex-col gap-2 p-3">
        <code>[&quot;{exec.join('", "')}&quot;]</code>
      </HoverCardContent>
    </HoverCard>
  ) : null;
};

export default PostStartExecs;
