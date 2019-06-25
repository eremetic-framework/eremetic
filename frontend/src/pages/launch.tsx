import React from 'react';

import TaskLauncher from '../components/TaskLauncher';
import HeaderBar from '../components/HeaderBar';

class Launch extends React.Component {
  render() {
    return (
      <div>
        <HeaderBar />
        <TaskLauncher />
      </div>
    );
  }
}

export default Launch;
