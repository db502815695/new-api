import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Empty,
  InputNumber,
  SideSheet,
  Space,
  Spin,
  Switch,
  Tag,
  Table,
  Typography,
} from '@douyinfe/semi-ui';
import { IconArrowLeft, IconRefresh } from '@douyinfe/semi-icons';
import { Link, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { API, isAdmin, showError, showSuccess } from '../../helpers';
import { GROUP_AVAILABILITY_ROUTE } from '../../components/layout/sidebarConfig';
import { useIsMobile } from '../../hooks/common/useIsMobile';

const { Title, Text } = Typography;

const STATUS_COLOR = {
  healthy: 'green',
  degraded: 'orange',
  down: 'red',
  unmonitored: 'grey',
};

const GroupAvailability = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const isMobile = useIsMobile();
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdated, setLastUpdated] = useState(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState('');
  const [channelRows, setChannelRows] = useState([]);
  const [channelLoading, setChannelLoading] = useState(false);
  const isConsoleRoute = location.pathname === GROUP_AVAILABILITY_ROUTE;
  const adminView = useMemo(() => isAdmin(), []);

  const statusLabel = useMemo(
    () => ({
      healthy: t('健康'),
      degraded: t('降级'),
      down: t('不可用'),
      unmonitored: t('未监控'),
    }),
    [t],
  );

  const formatPercent = useCallback((value) => {
    const num = Number(value || 0);
    return `${(num * 100).toFixed(1)}%`;
  }, []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/group_availability');
      if (res?.data?.success) {
        setGroups(res.data.data?.groups || []);
        setLastUpdated(new Date());
      } else {
        showError(res?.data?.message || t('加载失败'));
      }
    } catch (e) {
      showError(e?.message || t('加载失败'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  const fetchGroupChannels = useCallback(
    async (group) => {
      if (!adminView || !group) return;
      setChannelLoading(true);
      try {
        const res = await API.get('/api/channel_health/snapshots', {
          params: { group: group },
        });
        if (res?.data?.success) {
          setChannelRows(res.data.data || []);
        } else {
          showError(res?.data?.message || t('加载失败'));
        }
      } catch (e) {
        showError(e?.message || t('加载失败'));
      } finally {
        setChannelLoading(false);
      }
    },
    [adminView, t],
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (!autoRefresh) return undefined;
    const timer = setInterval(() => {
      fetchData();
      if (drawerVisible && selectedGroup) {
        fetchGroupChannels(selectedGroup);
      }
    }, 15000);
    return () => clearInterval(timer);
  }, [autoRefresh, drawerVisible, fetchData, fetchGroupChannels, selectedGroup]);

  const handleOpenDrawer = useCallback(
    async (group) => {
      if (!adminView) return;
      setSelectedGroup(group);
      setDrawerVisible(true);
      await fetchGroupChannels(group);
    },
    [adminView, fetchGroupChannels],
  );

  const handleSaveAbility = useCallback(
    async (channelId, field, value) => {
      try {
        const payload = { group: selectedGroup, channel_id: channelId };
        if (field === 'weight') {
          payload.weight = Math.max(0, Number(value));
        } else {
          payload.priority = Number(value);
        }
        const res = await API.post('/api/channel_health/ability', payload);
        if (res?.data?.success) {
          showSuccess(t('保存成功'));
          await fetchGroupChannels(selectedGroup);
        } else {
          showError(res?.data?.message || t('保存失败，请重试'));
        }
      } catch (e) {
        showError(e?.message || t('保存失败，请重试'));
      }
    },
    [fetchGroupChannels, selectedGroup, t],
  );

  const handleResetGroupHealth = useCallback(async () => {
    try {
      const res = await API.post('/api/channel_health/reset_group', {
        group: selectedGroup,
      });
      if (res?.data?.success) {
        showSuccess(t('重置成功'));
        await fetchData();
        await fetchGroupChannels(selectedGroup);
      } else {
        showError(res?.data?.message || t('重置失败，请重试'));
      }
    } catch (e) {
      showError(e?.message || t('重置失败，请重试'));
    }
  }, [fetchData, fetchGroupChannels, selectedGroup, t]);

  const channelColumns = useMemo(
    () => [
      {
        title: t('渠道名'),
        dataIndex: 'channel_name',
        key: 'channel_name',
      },
      {
        title: t('监控'),
        dataIndex: 'monitor_enabled',
        key: 'monitor_enabled',
        render: (value) =>
          value ? (
            <Tag color='green'>{t('已启用')}</Tag>
          ) : (
            <Tag color='grey'>{t('未启用')}</Tag>
          ),
      },
      {
        title: t('原始权重'),
        dataIndex: 'base_weight',
        key: 'base_weight',
      },
      {
        title: t('优先级'),
        dataIndex: 'priority',
        key: 'priority',
        render: (value, record) => (
          <InputNumber
            value={value}
            min={-1000}
            max={1000}
            style={{ width: 80 }}
            onBlur={(e) => {
              const newVal = Number(e.target.value);
              if (newVal !== value) {
                handleSaveAbility(record.channel_id, 'priority', newVal);
              }
            }}
            onEnterPress={(e) => {
              const newVal = Number(e.target.value);
              if (newVal !== value) {
                handleSaveAbility(record.channel_id, 'priority', newVal);
              }
            }}
          />
        ),
      },
      {
        title: t('权重'),
        dataIndex: 'weight',
        key: 'weight',
        render: (value, record) => (
          <InputNumber
            value={value}
            min={0}
            max={10000}
            style={{ width: 90 }}
            onBlur={(e) => {
              const newVal = Number(e.target.value);
              if (newVal !== value) {
                handleSaveAbility(record.channel_id, 'weight', newVal);
              }
            }}
            onEnterPress={(e) => {
              const newVal = Number(e.target.value);
              if (newVal !== value) {
                handleSaveAbility(record.channel_id, 'weight', newVal);
              }
            }}
          />
        ),
      },
      {
        title: t('逻辑权重'),
        dataIndex: 'effective_weight',
        key: 'effective_weight',
        render: (value, record) =>
          record.is_probe_weight ? (
            <span>
              {value} <Tag color='orange' size='small'>{t('探测')}</Tag>
            </span>
          ) : (
            value
          ),
      },
      {
        title: t('错误率'),
        dataIndex: 'error_rate',
        key: 'error_rate',
        render: (value) => formatPercent(value),
      },
      {
        title: t('平均首字耗时'),
        dataIndex: 'avg_ttft_ms',
        key: 'avg_ttft_ms',
        render: (value) => `${value || 0} ms`,
      },
      {
        title: t('样本'),
        dataIndex: 'requests',
        key: 'requests',
      },
      {
        title: t('错误数'),
        dataIndex: 'errors',
        key: 'errors',
      },
      {
        title: t('降权原因'),
        dataIndex: 'penalty_reason',
        key: 'penalty_reason',
        render: (value) => value || '-',
      },
    ],
    [formatPercent, handleSaveAbility, t],
  );

  const overviewMetrics = useMemo(() => {
    const monitoredGroups = groups.filter((g) => g.monitored);
    const healthyGroups = monitoredGroups.filter((g) => g.status === 'healthy');
    const avgAvailability = monitoredGroups.length
      ? monitoredGroups.reduce((sum, g) => sum + Number(g.availability || 0), 0) /
        monitoredGroups.length
      : 0;
    const avgTtft = monitoredGroups.length
      ? Math.round(
          monitoredGroups.reduce((sum, g) => sum + Number(g.avg_ttft_ms || 0), 0) /
            monitoredGroups.length,
        )
      : 0;

    return [
      {
        key: 'groups',
        label: t('监控分组'),
        value: `${monitoredGroups.length}/${groups.length}`,
        hint: t('启用监控'),
      },
      {
        key: 'healthy',
        label: t('健康分组'),
        value: `${healthyGroups.length}`,
        hint: t('健康状态'),
      },
      {
        key: 'availability',
        label: t('平均可用度'),
        value: `${(avgAvailability * 100).toFixed(1)}%`,
        hint: t('分组均值'),
      },
      {
        key: 'ttft',
        label: t('平均首字耗时'),
        value: `${avgTtft} ms`,
        hint: t('首字均值'),
      },
    ];
  }, [groups, t]);

  const selectedGroupSummary = useMemo(() => {
    const monitoredCount = channelRows.filter((row) => row.monitor_enabled).length;
    const totalRequests = channelRows.reduce(
      (sum, row) => sum + Number(row.requests || 0),
      0,
    );
    const avgWeightRatio =
      channelRows.length > 0
        ? channelRows.reduce((sum, row) => {
            const base = Number(row.base_weight || 0);
            const eff = Number(row.effective_weight || 0);
            if (base <= 0) return sum + 1;
            return sum + eff / base;
          }, 0) / channelRows.length
        : 0;

    return [
      {
        key: 'channels',
        label: t('渠道数'),
        value: channelRows.length,
      },
      {
        key: 'monitored',
        label: t('启用监控'),
        value: monitoredCount,
      },
      {
        key: 'requests',
        label: t('窗口样本'),
        value: totalRequests,
      },
      {
        key: 'weight',
        label: t('平均权重系数'),
        value: `${(avgWeightRatio * 100).toFixed(0)}%`,
      },
    ];
  }, [channelRows, t]);

  const renderCard = (g) => {
    const availability = g.monitored ? Math.round(g.availability * 100) : null;
    const color = STATUS_COLOR[g.status] || 'grey';
    return (
      <div className='group-availability-grid-item' key={g.group}>
        <Card
          className='group-availability-card group-availability-card-compact'
          shadows='hover'
          bodyStyle={{ padding: '18px 18px 16px' }}
          title={
            <div className='group-availability-card-head'>
              <Text strong className='group-availability-card-title'>
                {g.group || t('默认')}
              </Text>
              <Tag color={color} size='large'>
                {statusLabel[g.status] || g.status}
              </Tag>
            </div>
          }
        >
          {g.monitored ? (
            <div className='group-availability-card-body group-availability-card-grid'>
              <div className='group-availability-card-main'>
                <div className='group-availability-card-gauge'>
                  <div className='group-availability-card-percent'>
                    {availability}%
                  </div>
                  <Text
                    type='tertiary'
                    className='group-availability-card-subtitle'
                  >
                    {t('可用度')}
                  </Text>
                </div>
              </div>
              <div className='group-availability-card-bar-wrap'>
                <div className='group-availability-card-bar-track'>
                  <div
                    className='group-availability-card-bar'
                    style={{
                      width: `${availability}%`,
                      backgroundColor:
                        color === 'green'
                          ? 'var(--semi-color-success)'
                          : color === 'orange'
                            ? 'var(--semi-color-warning)'
                            : 'var(--semi-color-danger)',
                    }}
                  />
                </div>
              </div>
              <div className='group-availability-card-footer'>
                <div className='group-availability-card-stat'>
                  <Text type='tertiary'>{t('平均首字耗时')}</Text>
                  <div className='group-availability-card-stat-value'>
                    <Text strong>{g.avg_ttft_ms || 0}</Text>
                    <Text type='tertiary'> ms</Text>
                  </div>
                </div>
                <div className='group-availability-card-stat'>
                  <Text type='tertiary'>{t('健康渠道')}</Text>
                  <div className='group-availability-card-stat-value'>
                    <Text strong>{g.healthy_count}</Text>
                    <Text type='tertiary'> / {g.channel_count}</Text>
                  </div>
                </div>
              </div>
                {adminView && (
                  <div className='group-availability-card-actions'>
                    <Button
                      theme='light'
                      type='primary'
                      size='small'
                      onClick={() => handleOpenDrawer(g.group)}
                    >
                      {t('查看渠道')}
                    </Button>
                  </div>
                )}
            </div>
          ) : (
            <div className='group-availability-card-empty'>
              <Text type='tertiary'>
                {t('该分组下未开启监控的渠道，无法提供可用性数据')}
              </Text>
            </div>
          )}
        </Card>
      </div>
    );
  };

  return (
    <div
      className='group-availability-shell'
      style={
        isConsoleRoute
          ? { paddingTop: isMobile ? '52px' : '48px' }
          : { padding: '24px' }
      }
    >
      <div className='group-availability-head group-availability-head-compact'>
        <div className='group-availability-title-block'>
          {!isConsoleRoute && (
            <Link to='/usage-guide'>
              <Button icon={<IconArrowLeft />} theme='borderless'>
                {t('返回使用文档')}
              </Button>
            </Link>
          )}
          <div>
            <Title heading={3} className='group-availability-title'>
              {t('分组可用性监控')}
            </Title>
            <Text type='tertiary' className='group-availability-description'>
              {t(
                '基于真实流量成功率与首字耗时聚合分组状态，管理员可下钻查看组内渠道的逻辑权重变化。',
              )}
            </Text>
          </div>
        </div>
        <div className='group-availability-toolbar'>
          <Text type='tertiary'>
            {lastUpdated
              ? `${t('上次刷新')}: ${lastUpdated.toLocaleTimeString()}`
              : ''}
          </Text>
          <div className='group-availability-toolbar-actions'>
            <Text type='tertiary'>{t('自动刷新')}</Text>
            <Switch checked={autoRefresh} onChange={setAutoRefresh} />
            <Button icon={<IconRefresh />} onClick={fetchData} loading={loading}>
              {t('刷新')}
            </Button>
          </div>
        </div>
      </div>

      <div className='group-availability-overview'>
        {overviewMetrics.map((item) => (
          <div key={item.key} className='group-availability-metric'>
            <Text type='tertiary'>{item.label}</Text>
            <strong>{item.value}</strong>
            <small>{item.hint}</small>
          </div>
        ))}
      </div>

      <div className='group-availability-section-head'>
        <Text strong>{t('分组监控列表')}</Text>
        {adminView && <Text type='tertiary'>{t('支持管理员下钻查看渠道明细。')}</Text>}
      </div>

      <Spin spinning={loading && groups.length === 0}>
        {groups.length === 0 && !loading ? (
          <Empty
            title={t('暂无数据')}
            description={t(
              '请先在渠道编辑中开启"启用健康监控"，并在"运营设置-监控"中开启渠道健康度总开关',
            )}
          />
        ) : (
          <div className='group-availability-grid group-availability-grid-list group-availability-grid-cards group-availability-list-stack'>
            {groups.map(renderCard)}
          </div>
        )}
      </Spin>
      {adminView && (
        <SideSheet
          visible={drawerVisible}
          title={`${t('分组渠道健康度')} · ${selectedGroup || '-'}`}
          width={isMobile ? '100%' : 1040}
          onCancel={() => setDrawerVisible(false)}
          closeIcon={null}
          footer={
            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <Space>
                <Button
                  onClick={handleResetGroupHealth}
                  loading={channelLoading}
                >
                  {t('重置健康数据')}
                </Button>
                <Button
                  icon={<IconRefresh />}
                  onClick={() => fetchGroupChannels(selectedGroup)}
                  loading={channelLoading}
                >
                  {t('刷新')}
                </Button>
                <Button
                  theme='light'
                  type='primary'
                  onClick={() => setDrawerVisible(false)}
                >
                  {t('关闭')}
                </Button>
              </Space>
            </div>
          }
        >
          <Spin spinning={channelLoading}>
            <div className='group-availability-sheet-summary'>
              {selectedGroupSummary.map((item) => (
                <div key={item.key} className='group-availability-metric'>
                  <Text type='tertiary'>{item.label}</Text>
                  <strong>{item.value}</strong>
                </div>
              ))}
            </div>
            <Table
              rowKey='channel_id'
              columns={channelColumns}
              dataSource={channelRows}
              pagination={false}
              scroll={{ x: 860 }}
              empty={(
                <Empty
                  title={t('暂无数据')}
                  description={t('该分组当前没有可展示的渠道健康度数据')}
                />
              )}
            />
          </Spin>
        </SideSheet>
      )}
    </div>
  );
};

export default GroupAvailability;
