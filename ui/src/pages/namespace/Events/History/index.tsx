import { EventStreamingSubscriber } from "~/api/eventsv2/query/EventStreamingSubscriber";
import EventsList from "./EventsList";
import { FiltersObj } from "~/api/events/query/get";
import { useState } from "react";

export const itemsPerPage = 10;

const History = () => {
  const [filters, setFilters] = useState<FiltersObj>({});

  return (
    <>
      <EventsList filters={filters} setFilters={setFilters} />
      <EventStreamingSubscriber />
    </>
  );
};

export default History;
