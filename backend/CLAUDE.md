# LinkClaw Backend

## 数据库规范

### 禁止使用外键

数据库表之间**禁止使用外键约束**（FOREIGN KEY / REFERENCES）。原因：
- 外键的 `ON DELETE SET NULL` / `CASCADE` 会触发意料之外的级联副作用（如违反 CHECK 约束）
- 删除操作的顺序难以控制
- 关联查询和数据一致性由应用层（Go 代码）保证

编写迁移时：
- `CREATE TABLE` 中不写 `REFERENCES`
- 关联关系通过字段命名约定（`xxx_id`）+ 应用层 JOIN 查询实现
- 需要级联删除时在 Service 层手动处理
