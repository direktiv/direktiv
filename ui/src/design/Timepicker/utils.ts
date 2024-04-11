// this component is mostly copied from https://time.openstatus.dev/

/**
 * regular expression to check for valid hour format (01-23)
 */
function isValidHour(value: string) {
  return /^(0[0-9]|1[0-9]|2[0-3])$/.test(value);
}

/**
 * regular expression to check for valid minute format (00-59)
 */
function isValidMinuteOrSecond(value: string) {
  return /^[0-5][0-9]$/.test(value);
}

type GetValidNumberConfig = { max: number; min?: number; loop?: boolean };

function getValidNumber(
  value: string,
  { max, min = 0, loop = false }: GetValidNumberConfig
) {
  let numericValue = parseInt(value, 10);

  if (!isNaN(numericValue)) {
    if (!loop) {
      if (numericValue > max) numericValue = max;
      if (numericValue < min) numericValue = min;
    } else {
      if (numericValue > max) numericValue = min;
      if (numericValue < min) numericValue = max;
    }
    return numericValue.toString().padStart(2, "0");
  }

  return "00";
}

function getValidHour(value: string) {
  if (isValidHour(value)) return value;
  return getValidNumber(value, { max: 23 });
}

function getValidMinuteOrSecond(value: string) {
  if (isValidMinuteOrSecond(value)) return value;
  return getValidNumber(value, { max: 59 });
}

type GetValidIncrementNumberConfig = {
  min: number;
  max: number;
  increment: number;
};

function getValidIncrement(
  value: string,
  { min, max, increment }: GetValidIncrementNumberConfig
) {
  let numericValue = parseInt(value, 10);
  if (!isNaN(numericValue)) {
    numericValue += increment;
    return getValidNumber(String(numericValue), { min, max, loop: true });
  }
  return "00";
}

function getValidHourByIncrement(value: string, increment: number) {
  return getValidIncrement(value, { min: 0, max: 23, increment });
}

function getValidMinuteOrSecondByIncrement(value: string, increment: number) {
  return getValidIncrement(value, { min: 0, max: 59, increment });
}

function setMinutes(date: Date, value: string) {
  const minutes = getValidMinuteOrSecond(value);
  date.setMinutes(parseInt(minutes, 10));
  return date;
}

function setSeconds(date: Date, value: string) {
  const seconds = getValidMinuteOrSecond(value);
  date.setSeconds(parseInt(seconds, 10));
  return date;
}

function setHours(date: Date, value: string) {
  const hours = getValidHour(value);
  date.setHours(parseInt(hours, 10));
  return date;
}

export type TimePickerType = "minutes" | "seconds" | "hours";

export function updateDateByTime(
  date: Date,
  value: string,
  type: TimePickerType
) {
  switch (type) {
    case "minutes":
      return setMinutes(date, value);
    case "seconds":
      return setSeconds(date, value);
    case "hours":
      return setHours(date, value);
    default:
      return date;
  }
}

export function getTimeFromDate(date: Date, type: TimePickerType) {
  switch (type) {
    case "minutes":
      return getValidMinuteOrSecond(String(date.getMinutes()));
    case "seconds":
      return getValidMinuteOrSecond(String(date.getSeconds()));
    case "hours":
      return getValidHour(String(date.getHours()));
    default:
      return "00";
  }
}

export function getTimeByIncrementAndType(
  value: string,
  increment: number,
  type: TimePickerType
) {
  switch (type) {
    case "minutes":
    case "seconds":
      return getValidMinuteOrSecondByIncrement(value, increment);
    case "hours":
      return getValidHourByIncrement(value, increment);
    default:
      return "00";
  }
}
