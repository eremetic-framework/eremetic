import React from 'react';

import RunningTasks from '../components/RunningTasks';
import HeaderBar from '../components/HeaderBar';

class Index extends React.Component {
  render() {
    return (
      <div>
        <HeaderBar />
        <RunningTasks />
      </div>
    );
  }
}

export default Index;
