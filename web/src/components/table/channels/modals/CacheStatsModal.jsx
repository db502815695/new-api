/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useCallback, useEffect, useState } from 'react';
import {
  Modal,
  Button,
  RadioGroup,
  Radio,
  Table,
  Tag,
  Spin,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError } from '../../../../helpers';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

const CacheStatsModal = ({ visible, onCancel }) => {
  const { t } = useTranslation();
  const [days, setDays] = useState(7);
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState([]);

  const fetchData = useCallback(
    async (d) => {
      setLoading(true);
      try {
        const res = await API.get(`/api/channel/cache_stats?days=${d}`);
        if (res.data.success) {
          setData(res.data.data || []);
        } else {
          showError(res.data.message);
        }
      } catch (e) {
        showError(e.message);
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  useEffect(() => {
    if (visible) {
      fetchData(days);
    }
  }, [visible, days, fetchData]);

  const handleDaysChange = (e) => {
    setDays(e.target.value);
  };

  const columns = [
    {
      title: t('渠道名称'),
      dataIndex: 'channel_name',
      key: 'channel_name',
      render: (text) => <Text strong>{text || '-'}</Text>,
    },
    {
      title: t('总请求数'),
      dataIndex: 'total_requests',
      key: 'total_requests',
      sorter: (a, b) => a.total_requests - b.total_requests,
    },
    {
      title: t('缓存命中请求'),
      dataIndex: 'cache_hit_requests',
      key: 'cache_hit_requests',
    },
    {
      title: t('命中率'),
      dataIndex: 'hit_rate',
      key: 'hit_rate',
      sorter: (a, b) => a.hit_rate - b.hit_rate,
      render: (val) => {
        const pct = Math.round((val || 0) * 100);
        const color = pct >= 50 ? 'green' : 'orange';
        return <Tag color={color}>{pct}%</Tag>;
      },
    },
    {
      title: t('总 Prompt Tokens'),
      dataIndex: 'total_prompt_tokens',
      key: 'total_prompt_tokens',
    },
    {
      title: t('缓存命中 Tokens'),
      dataIndex: 'cache_hit_tokens',
      key: 'cache_hit_tokens',
    },
    {
      title: t('Token 命中率'),
      dataIndex: 'token_hit_ratio',
      key: 'token_hit_ratio',
      sorter: (a, b) => a.token_hit_ratio - b.token_hit_ratio,
      render: (val) => {
        const pct = Math.round((val || 0) * 100);
        const color = pct >= 50 ? 'green' : 'orange';
        return <Tag color={color}>{pct}%</Tag>;
      },
    },
    {
      title: t('缓存写入 Tokens'),
      dataIndex: 'cache_creation_tokens',
      key: 'cache_creation_tokens',
    },
  ];

  return (
    <Modal
      title={t('缓存命中率统计')}
      visible={visible}
      onCancel={onCancel}
      footer={
        <Button onClick={onCancel}>{t('关闭')}</Button>
      }
      width={1000}
      centered
    >
      <div style={{ marginBottom: 16 }}>
        <RadioGroup
          type='button'
          value={days}
          onChange={handleDaysChange}
        >
          <Radio value={1}>{t('近 1 天')}</Radio>
          <Radio value={7}>{t('近 7 天')}</Radio>
          <Radio value={30}>{t('近 30 天')}</Radio>
        </RadioGroup>
      </div>
      <Spin spinning={loading}>
        <Table
          columns={columns}
          dataSource={data}
          rowKey='channel_id'
          size='small'
          empty={t('暂无缓存统计数据')}
          pagination={false}
        />
      </Spin>
    </Modal>
  );
};

export default CacheStatsModal;
