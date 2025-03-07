#!/bin/bash

set -eu
# don't set any options, otherwise it will fail arbitrarily
# set -euo pipefail

echo_message() {
    echo "$1, $2, $3"
}

work_dir=$1
repo_url=$2
repo_name=$3
start_commit=$4
obsutil=$5 # the path of obsutil
obspath=$6 # obspath should has suffix of /

test -d $work_dir || mkdir -p $work_dir
cd $work_dir

# work_dir can't has suffix of / and must be an absolute path
work_dir=$(pwd)

git clone -q $repo_url
cd $repo_name

last_commit=$(git log --format="%H" -n 1)
file_prefix=$work_dir/$last_commit

all_files=${file_prefix}_files
if [ -z "$start_commit" ]; then
    rm .git -fr

    find . -type f > $all_files
    sed -i 's/^\.\///' $all_files
else
    git diff $start_commit..$last_commit --name-only > $all_files

    rm .git -fr
fi

if [ ! -s $all_files ]; then
    echo_message "$last_commit" "lfs" "no"

    exit 0
fi

lfs_files=${file_prefix}_lfs
small_files=${file_prefix}_small
deleted_files=${file_prefix}_deleted
while read line
do
    if [ -e "$line" ]; then
        sha=$(sed -n '/^oid sha256:.\{64\}$/p' $line)
        if [ -n "$sha" ]; then
            echo "$line:$sha" >> $lfs_files
        else
            echo $line >> $small_files
        fi
    else
        echo $line >> $deleted_files
    fi
done < $all_files

# handle small files
if [ -s $small_files ]; then
    n=$(wc -l $small_files | awk '{print $1}')
    find . -type f > $all_files
    other=$(wc -l $all_files | awk '{print $1}')
    other=$((other-n))
    sync_dir=""

    if [ $other -lt $n ]; then
        sync_dir=$(pwd)

        sed -i 's/^\.\///' $all_files

        while read line
        do
            if [ -z $(grep -Fx "$line" $small_files) ]; then
                rm $line
            fi
        done < $all_files
    else
        sync_dir=.git
        mkdir $sync_dir

        while read line
        do
            dir=$sync_dir/$(dirname $line)
            if [ ! -d $dir ]; then
                mkdir -p $dir
            fi

            mv $line $sync_dir/$line

        done < $small_files
    fi

    set +e

    $obsutil sync $sync_dir $obspath > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        success=0
        for i in {1..9}
        do
            sleep 0.5

            $obsutil sync $sync_dir $obspath > /dev/null 2>&1
            if [ $? -eq 0 ]; then
                success=1
                break
            fi
        done

        test $success -eq 1 || exit 1
    fi

    set -e
fi

set +e

if [ -s $deleted_files ]; then
    while read line
    do
        $obsutil rm ${obspath}$line -f > /dev/null 2>&1
    done < $deleted_files
fi

v="no"
test -s $lfs_files && v="yes"
echo_message "$last_commit" "$lfs_files" "$v"
