import React, { useEffect, useState } from "react";
import axios from "axios";

const App = () => {
  const [data, setData] = useState([]);

  // Функция для получения данных с API
  const fetchData = async () => {
    try {
      const response = await axios.get("http://localhost:8080/users"); // Укажите ваш адрес API
      const users = response.data;

      // Фильтруем данные, оставляя только последний пинг для каждого IP
      const latestPings = users.reduce((acc, user) => {
        if (
          !acc[user.ip] ||
          new Date(acc[user.ip].success_date || 0) < new Date(user.success_date || 0)
        ) {
          acc[user.ip] = user;
        }
        return acc;
      }, {});

      // Преобразуем объект обратно в массив
      setData(Object.values(latestPings));
    } catch (error) {
      console.error("Ошибка при получении данных:", error);
    }
  };

  // Загружаем данные при монтировании компонента
  useEffect(() => {
    fetchData();
  }, []);

  // Функция для форматирования длительности
  const formatDuration = (duration) => {
    if (duration === 0) return "Нет данных";
    const milliseconds = duration / 1e6; // Преобразуем наносекунды в миллисекунды
    return `${milliseconds.toFixed(2)} мс`;
  };

  return (
    <div style={{ padding: "20px" }}>
      <h1>Последние пинги</h1>
      <table border="1" cellPadding="10" cellSpacing="0">
        <thead>
          <tr>
            <th>IP-адрес</th>
            <th>Длительность пинга</th>
            <th>Время последней успешной попытки</th>
          </tr>
        </thead>
        <tbody>
          {data.map((user) => (
            <tr key={user.id}>
              <td>{user.ip}</td>
              <td>{formatDuration(user.duration)}</td>
              <td>
                {user.success_date
                  ? new Date(user.success_date).toLocaleString()
                  : "Нет данных"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default App;