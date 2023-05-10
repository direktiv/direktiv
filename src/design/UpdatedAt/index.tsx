import React, { FC, useCallback, useEffect, useRef } from "react";

import moment from "moment";

const useForceRerender = () => {
  const [, setState] = React.useState({ value: 10 });
  function rerenderForcefully() {
    setState((prev) => ({ ...prev }));
  }
  return rerenderForcefully;
};

const UpdatedAt: FC<{ date?: string }> = ({ date }) => {
  const forceUpdate = useForceRerender();

  const interval = useRef<ReturnType<typeof setInterval>>();

  const checkTime = useCallback(() => {
    const prev = moment(date);
    const now = moment(new Date());
    const duration = moment.duration(prev.diff(now));
    const mins = duration.asMinutes();
    if (mins < 60) {
      forceUpdate();
    } else {
      clearInterval(interval.current);
      forceUpdate();
    }
  }, [date, forceUpdate]);
  useEffect(() => {
    const prev = moment(date);
    const now = moment(new Date());
    const duration = moment.duration(now.diff(prev));
    const mins = duration.asMinutes();
    if (mins < 60) {
      interval.current = setInterval(() => {
        checkTime();
      }, 60000);
    }
    return () => {
      clearInterval(interval.current);
    };
  }, [date, checkTime]);
  return <>{moment(date).fromNow()}</>;
};

export default UpdatedAt;
