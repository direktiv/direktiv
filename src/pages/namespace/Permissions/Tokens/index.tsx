import { useTokens } from "~/api/enterprise/tokens/query/get";

const TokensPage = () => {
  const { data } = useTokens();

  return (
    <div>
      <h1>tokens</h1>
      {data?.tokens.map((token) => (
        <div key={token.id}>{token.description}</div>
      ))}
    </div>
  );
};

export default TokensPage;
