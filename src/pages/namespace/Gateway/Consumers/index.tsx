import { useConsumers } from "~/api/gateway/query/getConsumers";

const ConsumersPage = () => {
  const {
    data: gatewayList,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useConsumers();

  return <div>CONSUMERS</div>;
};

export default ConsumersPage;
