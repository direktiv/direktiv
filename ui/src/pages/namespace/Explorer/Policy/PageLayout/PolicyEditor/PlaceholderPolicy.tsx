import { AndGroup, Connector, OrGroup } from "~/design/Policy/Group";

import { Condition } from "~/design/Policy/Condition";
import { FC } from "react";
import { Placeholder } from "~/design/Policy/Placeholder";

export const PlaceholderPolicy: FC = () => (
  <OrGroup childSizes={[1]}>
    <AndGroup>
      <Condition>
        <div>user.email</div>
        <div>equal</div>
        <div>@example.org</div>
      </Condition>
      <Connector />
      <Placeholder />
    </AndGroup>
  </OrGroup>
);
