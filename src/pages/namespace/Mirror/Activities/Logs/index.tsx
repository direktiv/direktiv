import { pages } from "~/util/router/pages";

const Logs = () => {
  const { activity } = pages.mirror.useParams();

  return <div>{activity}</div>;
};

export default Logs;
