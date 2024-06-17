import { useEffect, useRef, useState } from "react";

import moment from "moment";

const minutesAgo = (date: moment.MomentInput) => {
  const prev = moment(date);
  const now = moment();
  return moment.duration(now.diff(prev)).asMinutes();
};

const useUpdatedAt = (
  date: moment.MomentInput,
  fallbackForInvalidDates = "-"
): string => {
  const isValidDate = moment(date).isValid();
  const [updatedString, setUpdatedString] = useState(
    moment(date).fromNow(true)
  );
  const [minAgo, setMinAgo] = useState(minutesAgo(date));
  const interval = useRef<ReturnType<typeof setInterval>>();

  useEffect(() => {
    /* must use fromNow(true) because otherwise after saving, it sometimes shows Updated in a few seconds */
    setUpdatedString(moment(date).fromNow(true));
    if (minAgo < 60) {
      interval.current = setInterval(() => {
        setMinAgo(minutesAgo(date));
      }, 60000);
    }
    return () => {
      clearInterval(interval.current);
    };
  }, [date, minAgo]);

  return isValidDate ? updatedString : fallbackForInvalidDates;
};

export default useUpdatedAt;
