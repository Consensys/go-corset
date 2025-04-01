;;error:7:9-10:expected list
;;error:7:11-12:expected list
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (A :i16) (B :i16))

(defconstraint c1 ()
  (let (C B)
    (if A
        (vanishes! C))))
