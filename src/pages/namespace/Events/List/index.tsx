import { FiltersObj, useEventsStream } from "~/api/events/query/get";

import EventsList from "./EventsList";
import { useState } from "react";

export const itemsPerPage = 15;

const EventsListWrapper = () => {
  const [filters, setFilters] = useState<FiltersObj>({});
  const [offset, setOffset] = useState(0);

  useEventsStream({ limit: itemsPerPage, offset, filters });

  return (
    <>
      <EventsList
        filters={filters}
        setFilters={setFilters}
        offset={offset}
        setOffset={setOffset}
      />
    </>
  );
};

export default EventsListWrapper;
