;;error:6:8-9:expected list
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (A :i16) (B :i16))

(defconstraint c1 ()
  (let C
    (if A
        (vanishes! C))))
