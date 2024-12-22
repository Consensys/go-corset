;;error:10:36-37:array index out-of-bounds
;;error:10:31-38:void expression not permitted here
(defcolumns
    (BIT :binary@prove :array [4])
    (ARG :i16@loob))

(defconstraint bits ()
  (- ARG
     (reduce +
      (for i [0:3] (* (^ 2 i) [BIT i])))))
