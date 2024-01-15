import ListenersList from "./ListenersList";
import { useEventListenersStream } from "~/api/eventListeners/query/get";
import { useState } from "react";

export const itemsPerPage = 10;

const Listeners = () => {
  const [offset, setOffset] = useState(0);

  useEventListenersStream({ limit: itemsPerPage, offset });

  return <ListenersList offset={offset} setOffset={setOffset} />;
};

export default Listeners;
