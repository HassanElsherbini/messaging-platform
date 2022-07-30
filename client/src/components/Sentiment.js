import React from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

const Sentiment = ({ dataset }) => {
  let data;
  if (dataset) {
    data = [
      {
        name: "sentiments",
        satisfied: dataset.satisfied,
        neutral: dataset.neutral,
        unsatisfied: dataset.unSatisified,
      },
    ];
  }
  return (
    <div className="dashboard-item sentiment">
      <h2>Sentiment</h2>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          width={500}
          height={300}
          data={data}
          margin={{
            top: 5,
            right: 30,
            left: 20,
            bottom: 5,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="sentiments" />
          <YAxis />
          <Tooltip />
          <Legend />
          <Bar dataKey="satisfied" fill="#009e80" />
          <Bar dataKey="neutral" fill="#70c6ec" />

          <Bar dataKey="unsatisfied" fill="#20232a" />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};

export default Sentiment;
