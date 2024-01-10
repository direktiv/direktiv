import Badge from "~/design/Badge";
import { FC } from "react";
import { RouteSchemaType } from "~/api/gateway/schema";

type AllowAnonymousProps = {
  methods: RouteSchemaType["methods"];
};

export const Methods: FC<AllowAnonymousProps> = ({ methods }) => (
  <div className="flex w-[190px] flex-wrap gap-1">
    {methods?.map((method) => (
      <Badge key={method} variant="outline">
        {method}
      </Badge>
    ))}
  </div>
);
