import { FC, useEffect, useRef, useState } from "react";

import moment from "moment";

const minutesAgo = (date: moment.MomentInput) => {
  const prev = moment(date);
  const now = moment();
  return moment.duration(now.diff(prev)).asMinutes();
};

const UpdatedAt: FC<{ date: moment.MomentInput }> = ({ date }) => {
  const [updatedString, setUpdatedString] = useState(moment(date).fromNow());
  const [minAgo, setMinAgo] = useState(minutesAgo(date));
  const interval = useRef<ReturnType<typeof setInterval>>();

  useEffect(() => {
    if (minAgo < 60) {
      interval.current = setInterval(() => {
        setMinAgo(minutesAgo(date));
        setUpdatedString(moment(date).fromNow());
      }, 60000);
    }
    return () => {
      clearInterval(interval.current);
    };
  }, [date, minAgo]);

  return <>{updatedString}</>;
};

export default UpdatedAt;
