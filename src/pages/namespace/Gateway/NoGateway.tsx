import { Card } from "~/design/Card";
import { Network } from "lucide-react";
import { NoResult } from "~/design/Table";
import { Trans } from "react-i18next";
import { gatewayNamespaceName } from "~/hooksNext/useIsGatewayAvailable";

const NoGateway = () => (
  <Card className="flex grow">
    <NoResult icon={Network}>
      <Trans
        i18nKey="pages.gateway.unavailable"
        values={{
          name: gatewayNamespaceName,
        }}
      />
    </NoResult>
  </Card>
);

export default NoGateway;
