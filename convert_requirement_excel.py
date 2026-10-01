"""
将行方需求清单 Excel 文件转换为系统导入格式的脚本
"""
import pandas as pd
from pathlib import Path
import sys

def convert_requirement_excel(input_file, output_file):
    """
    将原始需求清单 Excel 转换为系统导入格式
    
    目标格式列顺序：
    0. 需求名称
    1. 所属银行
    2. 需求编号
    3. 优先级 (HIGH/MEDIUM/LOW)
    4. 风险等级 (HIGH/MEDIUM/LOW)
    5. 状态 (PENDING_CLARIFY:待澄清/PENDING_SCHEDULING:待排期/DEVELOPING:开发中/TESTING:测试中/PENDING_RELEASE:待投产/COMPLETED:已完成)
    6. 排期 (yyyy-MM-dd HH:mm:ss)
    7. 主 R 系统
    8. 开发主 R|参与者
    9. 涉及系统
    10. 需求跟进人
    11. 提交人姓名
    12. 需求提出时间 (yyyy-MM-dd HH:mm:ss)
    13. 需求完成时间 (yyyy-MM-dd HH:mm:ss)
    14. 预计工作量（人天）
    15. 所属小组 (LOAN_MIDDLE:贷中/LOAN_PRE:贷前/RISK:风险)
    16. 最后修改人
    17. 需求描述
    18. 备注
    """
    
    print(f"读取源文件：{input_file}")
    
    try:
        # 读取 Excel 文件
        df_source = pd.read_excel(input_file)
        print(f"成功读取源文件，共 {len(df_source)} 行数据")
        print(f"源文件列名：{df_source.columns.tolist()}")
        
    except Exception as e:
        print(f"读取源文件失败：{e}")
        return False
    
    # 创建新的 DataFrame，按照目标格式重新排列
    df_target = pd.DataFrame()
    
    # 列映射关系（根据实际源文件格式调整）
    # 这里需要根据实际的 Excel 列名进行映射
    column_mapping = {
        '需求名称': '需求名称',
        '需求标题': '需求名称',
        'title': '需求名称',
        '所属银行': '所属银行',
        'bankName': '所属银行',
        '银行': '所属银行',
        '需求编号': '需求编号',
        'requirementNo': '需求编号',
        '编号': '需求编号',
        '优先级': '优先级 (HIGH/MEDIUM/LOW)',
        'priority': '优先级 (HIGH/MEDIUM/LOW)',
        '风险等级': '风险等级 (HIGH/MEDIUM/LOW)',
        'riskLevel': '风险等级 (HIGH/MEDIUM/LOW)',
        '状态': '状态 (PENDING_CLARIFY:待澄清/PENDING_SCHEDULING:待排期/DEVELOPING:开发中/TESTING:测试中/PENDING_RELEASE:待投产/COMPLETED:已完成)',
        'status': '状态 (PENDING_CLARIFY:待澄清/PENDING_SCHEDULING:待排期/DEVELOPING:开发中/TESTING:测试中/PENDING_RELEASE:待投产/COMPLETED:已完成)',
        '排期': '排期 (yyyy-MM-dd HH:mm:ss)',
        'scheduleDate': '排期 (yyyy-MM-dd HH:mm:ss)',
        '计划日期': '排期 (yyyy-MM-dd HH:mm:ss)',
        '主 R 系统': '主 R 系统',
        'mainSystem': '主 R 系统',
        '开发主 R|参与者': '开发主 R|参与者',
        'developer': '开发主 R|参与者',
        '开发人员': '开发主 R|参与者',
        '涉及系统': '涉及系统',
        'involvedSystems': '涉及系统',
        '需求跟进人': '需求跟进人',
        'followerName': '需求跟进人',
        '跟进人': '需求跟进人',
        '负责人': '需求跟进人',
        '提交人姓名': '提交人姓名',
        'submitterName': '提交人姓名',
        '提交人': '提交人姓名',
        '需求提出时间': '需求提出时间 (yyyy-MM-dd HH:mm:ss)',
        'submitTime': '需求提出时间 (yyyy-MM-dd HH:mm:ss)',
        '提出时间': '需求提出时间 (yyyy-MM-dd HH:mm:ss)',
        '需求完成时间': '需求完成时间 (yyyy-MM-dd HH:mm:ss)',
        'completionDate': '需求完成时间 (yyyy-MM-dd HH:mm:ss)',
        '完成时间': '需求完成时间 (yyyy-MM-dd HH:mm:ss)',
        '预计工作量（人天）': '预计工作量（人天）',
        'estimatedWorkload': '预计工作量（人天）',
        '工作量': '预计工作量（人天）',
        '所属小组': '所属小组 (LOAN_MIDDLE:贷中/LOAN_PRE:贷前/RISK:风险)',
        'groupType': '所属小组 (LOAN_MIDDLE:贷中/LOAN_PRE:贷前/RISK:风险)',
        '小组': '所属小组 (LOAN_MIDDLE:贷中/LOAN_PRE:贷前/RISK:风险)',
        '最后修改人': '最后修改人',
        'lastModifierName': '最后修改人',
        '修改人': '最后修改人',
        '需求描述': '需求描述',
        'description': '需求描述',
        '描述': '需求描述',
        '备注': '备注',
        'remark': '备注',
        '备注说明': '备注',
    }
    
    # 目标列（按顺序）
    target_columns = [
        '需求名称',
        '所属银行',
        '需求编号',
        '优先级 (HIGH/MEDIUM/LOW)',
        '风险等级 (HIGH/MEDIUM/LOW)',
        '状态 (PENDING_CLARIFY:待澄清/PENDING_SCHEDULING:待排期/DEVELOPING:开发中/TESTING:测试中/PENDING_RELEASE:待投产/COMPLETED:已完成)',
        '排期 (yyyy-MM-dd HH:mm:ss)',
        '主 R 系统',
        '开发主 R|参与者',
        '涉及系统',
        '需求跟进人',
        '提交人姓名',
        '需求提出时间 (yyyy-MM-dd HH:mm:ss)',
        '需求完成时间 (yyyy-MM-dd HH:mm:ss)',
        '预计工作量（人天）',
        '所属小组 (LOAN_MIDDLE:贷中/LOAN_PRE:贷前/RISK:风险)',
        '最后修改人',
        '需求描述',
        '备注'
    ]
    
    # 尝试映射每一列
    for target_col in target_columns:
        found = False
        
        # 先精确匹配
        if target_col in df_source.columns:
            df_target[target_col] = df_source[target_col]
            found = True
        else:
            # 模糊匹配可能的列名
            for source_col in df_source.columns:
                # 检查是否在映射关系中
                if source_col in column_mapping and column_mapping[source_col] == target_col:
                    df_target[target_col] = df_source[source_col]
                    found = True
                    break
                
                # 检查是否包含关键词
                if any(keyword in source_col for keyword in get_keywords(target_col)):
                    df_target[target_col] = df_source[source_col]
                    found = True
                    break
        
        # 如果没找到，创建空列
        if not found:
            df_target[target_col] = ''
    
    print(f"\n目标文件列名：{df_target.columns.tolist()}")
    print(f"目标文件行数：{len(df_target)}")
    
    # 导出到新 Excel
    try:
        # 使用 xlsxwriter 引擎来更好地控制格式
        writer = pd.ExcelWriter(output_file, engine='xlsxwriter')
        df_target.to_excel(writer, sheet_name='需求导入模板', index=False)
        
        # 获取 workbook 和 worksheet
        workbook = writer.book
        worksheet = writer.sheets['需求导入模板']
        
        # 设置列宽
        column_widths = [25, 20, 20, 20, 20, 40, 25, 20, 25, 30, 20, 20, 25, 25, 20, 30, 20, 40, 30]
        for i, width in enumerate(column_widths):
            col_letter = chr(65 + i)  # A, B, C, ...
            worksheet.set_column(f'{col_letter}:{col_letter}', width)
        
        # 设置表头行高
        worksheet.set_row(0, 20)
        
        writer.close()
        print(f"\n成功生成目标文件：{output_file}")
        return True
        
    except Exception as e:
        print(f"生成目标文件失败：{e}")
        return False


def get_keywords(column_name):
    """从列名中提取关键词用于模糊匹配"""
    # 去除括号内的说明文字
    base_name = column_name.split('(')[0].split('（')[0]
    return [base_name.strip()]


def main():
    # 文件路径
    script_dir = Path(__file__).parent
    parent_dir = script_dir.parent
    
    # 输入文件路径（sql/data 目录下的原始文件）
    input_file = parent_dir / 'sql' / 'data' / '(行方需求清单) 行方需求清单.xlsx'
    
    # 输出文件路径（sql/data 目录下生成的转换后文件）
    output_file = parent_dir / 'sql' / 'data' / '需求导入数据_转换后.xlsx'
    
    print("=" * 60)
    print("行方需求清单 Excel 转换工具")
    print("=" * 60)
    print(f"\n源文件：{input_file}")
    print(f"目标文件：{output_file}")
    print("\n开始转换...\n")
    
    # 检查源文件是否存在
    if not input_file.exists():
        print(f"错误：源文件不存在：{input_file}")
        sys.exit(1)
    
    # 执行转换
    success = convert_requirement_excel(input_file, output_file)
    
    if success:
        print("\n" + "=" * 60)
        print("转换完成！")
        print("=" * 60)
        print(f"\n生成的文件已保存至：{output_file}")
        print("\n您可以使用此文件在系统中进行批量导入操作。")
    else:
        print("\n转换失败，请检查错误信息。")
        sys.exit(1)


if __name__ == '__main__':
    main()
