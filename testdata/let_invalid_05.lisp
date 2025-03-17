;;error:6:16-17:already defined
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (A :i16) (B :i16))

(defconstraint c1 ()
  (let ((C B) (C B))
    (if A
        (vanishes! C))))
