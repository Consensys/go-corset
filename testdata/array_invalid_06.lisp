;;error:10:36-37:array index out-of-bounds
;;
(defcolumns
    (BIT :binary@prove :array [4])
    (ARG :i16))

(defconstraint bits ()
  (== ARG
     (reduce +
      (for i [0:3] (* (^ 2 i) [BIT i])))))
