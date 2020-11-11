# ufr

Remember this?:

```fish
for f in (cat file_list.txt)
    aws s3 mv s3://<bucket>/$f s3://<bucket>/<actual-intended-parent-prefix>/
end
```

Never again!
