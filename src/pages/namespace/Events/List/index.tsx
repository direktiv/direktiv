import { FiltersObj, useEventsStream } from "~/api/events/query/get";

import EventsList from "./EventsList";
import { useState } from "react";

const itemsPerPage = 15;

const EventsListWrapper = () => {
  const [filters, setFilters] = useState<FiltersObj>({});

  useEventsStream({ limit: itemsPerPage, offset: 0, filters });

  return <EventsList filters={filters} setFilters={setFilters} />;
};

export default EventsListWrapper;
