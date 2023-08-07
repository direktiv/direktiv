import { useEvents } from "~/api/events/query/get";

const eventsPerPage = 15;

const EventsPageList = () => {
  const { data, isFetched } = useEvents({
    limit: eventsPerPage,
    offset: 0,
    filters: {},
  });

  return (
    <div>
      <ul>
        {data?.events.results.map((event, i) => (
          <li key={i}>{event.type}</li>
        ))}
      </ul>
    </div>
  );
};

export default EventsPageList;
