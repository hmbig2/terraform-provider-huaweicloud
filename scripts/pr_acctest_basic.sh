#!/bin/bash

all_files=$(git diff $1 $2 --name-only huaweicloud | grep -v "_test.go")

for f in $all_files; do

    path=${f%/*}
    if [ "$path" != "huaweicloud" ]; then
        # update path to "huaweicloud/services/acceptance/xxx"
        path=${path/services/services\/acceptance}
    fi

    org_file=${f##*/}
    test_file=$path/${org_file/%.go/_test.go}

    if [ -f "./${test_file}" ]; then
        basic_cases=$(grep "^func TestAcc" ./${test_file} | grep _basic | awk '{print $2}' | awk -F '(' '{print $1}')
        if [ "X$basic_cases" != "X" ]; then
            echo -e "\n\`\`\` \nrun acceptance basic tests: $basic_cases"
            make testacc TEST="./$path" TESTARGS="-run ${basic_cases}"
            echo "\`\`\`"
        fi
    else
        echo -e "**[skipped]** --- ./${test_file} does not exist"
    fi
done
