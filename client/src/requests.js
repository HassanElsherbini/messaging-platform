import axios from "axios";

const FetchAnalytics = async () => {
  try {
    const result = await axios.get("/api/analytics/");
    return result.data;
  } catch (err) {
    console.log(err);
  }
};

export { FetchAnalytics };
