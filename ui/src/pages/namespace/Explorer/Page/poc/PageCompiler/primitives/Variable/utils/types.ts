export type Result<T, E> = Success<T> | Failure<E>;

type Success<T> = {
  success: true;
  data: T;
};

type Failure<E> = {
  success: false;
  error: E;
};
