import { EventStreamingSubscriber } from "~/api/events/query/EventStreamingSubscriber";
import EventsList from "./EventsList";
import { FiltersSchemaType } from "~/api/events/schema/filters";
import { useState } from "react";

const itemsPerPage = 10;

const History = () => {
  const [filters, setFilters] = useState<FiltersSchemaType>({});

  return (
    <>
      <EventsList filters={filters} setFilters={setFilters} />
      <EventStreamingSubscriber filters={filters} />
    </>
  );
};

export default History;
