;;error:12:12-18:out-of-bounds array access
(defcolumns
    (BIT :binary@prove :array [4])
    (ARG :i16@loob))

(defconstraint bits ()
  (- ARG
     (+
      (* 1 [BIT 0])
      (* 2 [BIT 1])
      (* 4 [BIT 2])
      (* 8 [BIT 3]))))
