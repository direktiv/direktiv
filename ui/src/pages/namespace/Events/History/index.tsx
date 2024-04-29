import { EventStreamingSubscriber } from "~/api/eventsv2/query/EventStreamingSubscriber";
import EventsList from "./EventsList";
import { FiltersObj } from "~/api/events/query/get";
import { useState } from "react";

export const itemsPerPage = 10;

const History = () => {
  const [filters, setFilters] = useState<FiltersObj>({});
  const [offset, setOffset] = useState(0);

  // useEventsStream({ limit: itemsPerPage, offset, filters });

  return (
    <>
      <EventsList
        filters={filters}
        setFilters={setFilters}
        offset={offset}
        setOffset={setOffset}
      />
      <EventStreamingSubscriber />
    </>
  );
};

export default History;
