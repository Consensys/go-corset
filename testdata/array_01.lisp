(defcolumns
    (BIT :binary@prove :array [4])
    (ARG :i16@loob))

(defconstraint bits ()
  (- ARG
     (+
      (* 1 [BIT 1])
      (* 2 [BIT 2])
      (* 4 [BIT 3])
      (* 8 [BIT 4]))))
