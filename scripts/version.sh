#!/bin/bash

# 版本管理辅助脚本
# 用法: ./scripts/version.sh [命令] [参数]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 显示帮助信息
show_help() {
    echo "版本管理辅助脚本"
    echo ""
    echo "用法: $0 [命令] [参数]"
    echo ""
    echo "命令:"
    echo "  current           显示当前版本"
    echo "  next [类型]       计算下一个版本号"
    echo "  validate <版本>   验证版本号格式"
    echo "  compare <v1> <v2> 比较两个版本号"
    echo "  list              列出所有版本标签"
    echo "  help              显示此帮助信息"
    echo ""
    echo "版本类型 (用于 next 命令):"
    echo "  major             主版本号 +1"
    echo "  minor             次版本号 +1"
    echo "  patch             修订号 +1 (默认)"
    echo "  alpha             添加 alpha 预发布标识"
    echo "  beta              添加 beta 预发布标识"
    echo "  rc                添加 rc 预发布标识"
    echo ""
    echo "示例:"
    echo "  $0 current"
    echo "  $0 next minor"
    echo "  $0 validate 1.2.3"
    echo "  $0 compare 1.2.3 1.2.4"
}

# 获取当前版本
get_current_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$latest_tag" ]; then
        echo "0.0.0"
    else
        echo "${latest_tag#v}"
    fi
}

# 验证版本号格式
validate_version() {
    local version=$1
    if [[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+(\.[0-9]+)?)?$ ]]; then
        return 0
    else
        return 1
    fi
}

# 解析版本号
parse_version() {
    local version=$1
    local major minor patch prerelease
    
    # 移除 v 前缀
    version=${version#v}
    
    # 分离预发布标识
    if [[ $version == *-* ]]; then
        prerelease=${version#*-}
        version=${version%-*}
    fi
    
    # 分离主版本、次版本、修订号
    IFS='.' read -r major minor patch <<< "$version"
    
    echo "$major $minor $patch $prerelease"
}

# 计算下一个版本号
calculate_next_version() {
    local current_version=$1
    local bump_type=${2:-patch}
    
    local parsed=($(parse_version "$current_version"))
    local major=${parsed[0]}
    local minor=${parsed[1]}
    local patch=${parsed[2]}
    local prerelease=${parsed[3]}
    
    case $bump_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            prerelease=""
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            prerelease=""
            ;;
        patch)
            patch=$((patch + 1))
            prerelease=""
            ;;
        alpha)
            if [ -z "$prerelease" ]; then
                patch=$((patch + 1))
                prerelease="alpha.1"
            elif [[ $prerelease == alpha.* ]]; then
                local alpha_num=${prerelease#alpha.}
                prerelease="alpha.$((alpha_num + 1))"
            else
                patch=$((patch + 1))
                prerelease="alpha.1"
            fi
            ;;
        beta)
            if [ -z "$prerelease" ]; then
                patch=$((patch + 1))
                prerelease="beta.1"
            elif [[ $prerelease == beta.* ]]; then
                local beta_num=${prerelease#beta.}
                prerelease="beta.$((beta_num + 1))"
            else
                patch=$((patch + 1))
                prerelease="beta.1"
            fi
            ;;
        rc)
            if [ -z "$prerelease" ]; then
                patch=$((patch + 1))
                prerelease="rc.1"
            elif [[ $prerelease == rc.* ]]; then
                local rc_num=${prerelease#rc.}
                prerelease="rc.$((rc_num + 1))"
            else
                patch=$((patch + 1))
                prerelease="rc.1"
            fi
            ;;
        *)
            print_message $RED "错误: 未知的版本类型 '$bump_type'"
            exit 1
            ;;
    esac
    
    if [ -n "$prerelease" ]; then
        echo "$major.$minor.$patch-$prerelease"
    else
        echo "$major.$minor.$patch"
    fi
}

# 比较版本号
compare_versions() {
    local v1=$1
    local v2=$2
    
    # 移除 v 前缀
    v1=${v1#v}
    v2=${v2#v}
    
    if [ "$v1" = "$v2" ]; then
        echo "equal"
        return 0
    fi
    
    # 使用 sort -V 进行版本比较
    local sorted=$(printf '%s\n%s\n' "$v1" "$v2" | sort -V)
    local first_line=$(echo "$sorted" | head -n1)
    
    if [ "$first_line" = "$v1" ]; then
        echo "less"
    else
        echo "greater"
    fi
}

# 列出所有版本标签
list_versions() {
    print_message $BLUE "所有版本标签:"
    git tag -l "v*" --sort=-version:refname | head -20
    
    local total_count=$(git tag -l "v*" | wc -l)
    if [ $total_count -gt 20 ]; then
        print_message $YELLOW "... 还有 $((total_count - 20)) 个版本 (使用 'git tag -l \"v*\"' 查看全部)"
    fi
}

# 主函数
main() {
    local command=${1:-help}
    
    case $command in
        current)
            local current=$(get_current_version)
            print_message $GREEN "当前版本: $current"
            ;;
        next)
            local current=$(get_current_version)
            local bump_type=${2:-patch}
            local next=$(calculate_next_version "$current" "$bump_type")
            print_message $BLUE "当前版本: $current"
            print_message $GREEN "下一个版本 ($bump_type): $next"
            ;;
        validate)
            local version=$2
            if [ -z "$version" ]; then
                print_message $RED "错误: 请提供要验证的版本号"
                exit 1
            fi
            if validate_version "$version"; then
                print_message $GREEN "版本号 '$version' 格式正确"
            else
                print_message $RED "版本号 '$version' 格式不正确"
                exit 1
            fi
            ;;
        compare)
            local v1=$2
            local v2=$3
            if [ -z "$v1" ] || [ -z "$v2" ]; then
                print_message $RED "错误: 请提供两个版本号进行比较"
                exit 1
            fi
            local result=$(compare_versions "$v1" "$v2")
            case $result in
                equal)
                    print_message $BLUE "$v1 = $v2"
                    ;;
                less)
                    print_message $BLUE "$v1 < $v2"
                    ;;
                greater)
                    print_message $BLUE "$v1 > $v2"
                    ;;
            esac
            ;;
        list)
            list_versions
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_message $RED "错误: 未知命令 '$command'"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 检查是否在 Git 仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_message $RED "错误: 当前目录不是 Git 仓库"
    exit 1
fi

main "$@"