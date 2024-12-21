(defpurefun ((eq! :@loob) x y) (- x y))
;;
(defcolumns
  (STAMP :i32)
  (MXP_TYPE :binary@prove :array [5]))

(defconstraint type-flag-sum (:guard STAMP)
  (eq! 1
       (reduce + (for i [5] [MXP_TYPE i]))))
