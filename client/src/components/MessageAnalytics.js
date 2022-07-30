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

const MessageAnalytics = ({ dataset }) => {
  let data;
  if (dataset) {
    data = transformData(dataset);
  }
  return (
    <div className="dashboard-item analytics">
      <h2>Message Analytics</h2>
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
          <XAxis dataKey="name" />
          <YAxis />
          <Tooltip />
          <Legend />
          <Bar dataKey="sent" fill="#009e80" />
          <Bar dataKey="read" fill="#70c6ec" />

          <Bar dataKey="response" fill="#20232a" />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};

const transformData = (analytics) => {
  const data = new Array(7).fill(null);
  const mapName = {
    monday: { name: "MON", key: 0 },
    tuesday: { name: "TUE", key: 1 },
    wednesday: { name: "WED", key: 2 },
    thursday: { name: "THU", key: 3 },
    friday: { name: "FRI", key: 4 },
    saturday: { name: "SAT", key: 5 },
    sunday: { name: "SUN", key: 6 },
  };

  for (const key in analytics) {
    const val = analytics[key];
    const mapping = mapName[key];

    data[mapping.key] = {
      name: mapping.name,
      read: val.read,
      sent: val.sent,
      response: val.replied,
    };
  }

  return data;
};

export default MessageAnalytics;
