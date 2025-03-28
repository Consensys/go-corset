(defcolumns
  (STAMP :i32)
  (MXP_TYPE :binary@prove :array [5]))

(defconstraint type-flag-sum (:guard STAMP)
  (== 1
       (reduce + (for i [5] [MXP_TYPE i]))))
