import http from "k6/http";

export const options = {
  scenarios: {
    frontpage: {
      executor: "shared-iterations",
      vus: 1000,
      iterations: 10000,
    },
  },
};

const baseUrl = "https://orange.decode.ee";
export default function () {
  let response = http.get(baseUrl);
  let doc = response.html();
  doc
    .find("script")
    .toArray()
    .forEach((script) => {
      http.get(baseUrl + script.attr("src"));
    });
}
