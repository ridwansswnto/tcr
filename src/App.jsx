import React, {useEffect, useState} from "react";

function App(){
  const [jobs, setJobs] = useState([]);
  const [runners, setRunners] = useState({});

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const j = await fetch("/jobs").then(r => r.json());
        const rns = await fetch("/runners").then(r => r.json());
        setJobs(j);
        setRunners(rns);
      } catch (e) {
        console.error(e);
      }
    };
    fetchAll();
    const id = setInterval(fetchAll, 5000);
    return () => clearInterval(id);
  }, []);

  return (
    <div style={{padding:20, fontFamily:"Inter, sans-serif"}}>
      <h1>TCR Dashboard</h1>

      <section>
        <h2>Runners</h2>
        <table border="1" cellPadding="8">
          <thead><tr><th>ID</th><th>Address</th><th>Port</th><th>LastSeen</th><th>IsBusy</th></tr></thead>
          <tbody>
            {Object.values(runners).map(r => (
              <tr key={r.ID}>
                <td>{r.ID}</td>
                <td>{r.Address}</td>
                <td>{r.Port}</td>
                <td>{r.LastSeen}</td>
                <td>{r.IsBusy ? "yes" : "no"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <section style={{marginTop:20}}>
        <h2>Jobs</h2>
        <table border="1" cellPadding="8">
          <thead><tr><th>ID</th><th>Repo</th><th>Job</th><th>Status</th><th>Created</th></tr></thead>
          <tbody>
            {jobs.map(j => (
              <tr key={j.id}>
                <td>{j.id}</td>
                <td>{j.repo_owner}/{j.repo_name}</td>
                <td>{j.job_name}</td>
                <td>{j.status}</td>
                <td>{j.created_at}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </div>
  );
}

export default App;
