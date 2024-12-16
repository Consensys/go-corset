(defcolumns
    (BIT :binary@prove :array [4])
    (ARG :i16@loob))

(defconstraint bits ()
  (- ARG
     (reduce +
      (for i [0:3] (* (^ 2 i) [BIT (+ 1 i)])))))
