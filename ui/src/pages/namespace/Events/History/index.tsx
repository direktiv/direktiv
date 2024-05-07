import { EventStreamingSubscriber } from "~/api/eventsv2/query/EventStreamingSubscriber";
import EventsList from "./EventsList";
import { FiltersObj } from "~/api/events/query/get";
import { useState } from "react";

export const itemsPerPage = 10;

const History = () => {
  // temporarily hard coded the filters - pending re-implementation
  const [_, setFilters] = useState<FiltersObj>({});
  const filters = { typeContains: "foo" };

  return (
    <>
      <EventsList filters={filters} setFilters={setFilters} />
      <EventStreamingSubscriber filters={filters} />
    </>
  );
};

export default History;
