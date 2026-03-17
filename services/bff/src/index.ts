import Fastify from "fastify";

const port = Number(process.env.PORT) || 3000;

const app = Fastify({ logger: true });

app.get("/health", async () => {
  return { status: "ok" };
});

app.listen({ port, host: "0.0.0.0" }, (err) => {
  if (err) {
    app.log.error(err);
    process.exit(1);
  }
});
