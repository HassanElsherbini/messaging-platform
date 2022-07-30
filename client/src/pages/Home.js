import React, { useEffect, useState } from "react";
import MessageAnalytics from "../components/MessageAnalytics";
import Sentiment from "../components/Sentiment";
import "./home.css";

import { FetchAnalytics } from "../requests";

export default function Home() {
  const [analytics, setAnalytics] = useState({});
  useEffect(() => {
    (async () => {
      const data = await FetchAnalytics();
      setAnalytics(data);
    })();
  }, []);

  return (
    <div className="home-container">
      <section className="sidebar" />
      <section className="dashboard-container">
        <section className="dashboard-content">
          <MessageAnalytics dataset={analytics.byDay} />
          <Sentiment dataset={analytics.sentiment} />
        </section>
      </section>
    </div>
  );
}
