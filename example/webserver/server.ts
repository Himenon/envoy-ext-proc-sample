import * as expres from "express";
import * as compression from "compression";

const app = expres();

const PORT: number = parseInt(process.env.PORT, 10) || 3000;

app.use(
  compression({
    threshold: 0,
  })
);

app.use((req, res, next) => {
  const message = JSON.stringify({ Method: req.method, Path: req.url });
  console.log(message);
  next();
});

app.get("/", (req, res) => {
  res.send("<h1>Hello world!</h1>");
  res.end();
});

app.get("/error", (req, res) => {
  const queryStatusCode = req.query["status"];
  const statusCode: number =
    typeof queryStatusCode === "string" ? parseInt(queryStatusCode, 10) : 200;
  console.log(JSON.stringify({ StatusCode: statusCode }));
  res.status(statusCode);
  res.end();
});


app.listen(PORT, () => {
  console.log(`Start server: http://localhost:${PORT}`);
});
