
;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; Columns
;;;;;;;;;;;;;;;;;;;;;;;;;;
;; TODO: does not accept field type
(defcolumns
    (BOXES :i3 :array [9])
    (STAMP :i1)
)

;;X is 1
;;O is 2

;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; Shorthands
;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun (sum-boxes)
       (reduce + (for i [9] [BOXES i]))
)

;; (defun (sum-boxes-test)
;;       (reduce + (for i [1:9:1] [BOXES i]))
;;)

(defun (sum-boxes-next)
       (reduce + (for i [9] (next [BOXES i])))
)

(defun (multiply-boxes)
       (reduce * (for i [9] [BOXES i]))
)

(defun (diff-all)
       (reduce + (for i [9] (- [BOXES i] (prev [BOXES i]))))
)

(defun (diff-all-next)
       (reduce + (for i [9] (- (next [BOXES i]) [BOXES i])))
)

(defun (sum-column-1)
 (+ (+ [BOXES 1] [BOXES 4]) [BOXES 7])
)

(defun (sum-column-2)
 (+ (+ [BOXES 2] [BOXES 5]) [BOXES 8])
)

(defun (sum-column-3)
 (+ (+ [BOXES 3] [BOXES 6]) [BOXES 9])
)

(defun (sum-row-1)
       (reduce + (for i [1:3] [BOXES i]))
)

(defun (sum-row-2)
       (reduce + (for i [4:6] [BOXES i]))
)

(defun (sum-row-3)
       (reduce + (for i [7:9] [BOXES i]))
)

(defun (sum-diagonal-left-right)
 (+ (+ [BOXES 1] [BOXES 5]) [BOXES 9])
)

(defun (sum-diagonal-right-left)
 (+ (+ [BOXES 3] [BOXES 5]) [BOXES 7])
)

(defun (multiply-column-1)
 (* (* [BOXES 1] [BOXES 4]) [BOXES 7])
)

(defun (multiply-column-2)
 (* (* [BOXES 2] [BOXES 5]) [BOXES 8])
)

(defun (multiply-column-3)
 (* (* [BOXES 3] [BOXES 6]) [BOXES 9])
)

(defun (multiply-row-1)
       (reduce * (for i [1:3] [BOXES i]))
)

(defun (multiply-row-2)
       (reduce * (for i [4:6] [BOXES i]))
)

(defun (multiply-row-3)
       (reduce * (for i [7:9] [BOXES i]))
)

(defun (multiply-diagonal-left-right)
 (* (* [BOXES 1] [BOXES 5]) [BOXES 9])
)

(defun (multiply-diagonal-right-left)
 (* (* [BOXES 3] [BOXES 5]) [BOXES 7])
)

;; check if the game is won or finished (0)
;; game is won if sum is 3 and multiplication is 1 (to discard [0,1,2] combination)
;; game is won if sum is 6
;; TODO: and! not working with command line
(defun (no-one-has-won-yet-or-finished-or-draw-version1)
        (if (and! (eq! 3 (sum-column-1)) (eq! 1 (multiply-column-1))) 0
           (if (eq! 6 (sum-column-1)) 0
        (if (and! (eq! 3 (sum-column-2)) (eq! 1 (multiply-column-2))) 0
            (if (eq! 6 (sum-column-2)) 0
        (if (and! (eq! 3 (sum-column-3)) (eq! 1 (multiply-column-3))) 0
            (if (eq! 6 (sum-column-3)) 0
        (if (and! (eq! 3 (sum-row-1)) (eq! 1 (multiply-row-1))) 0
            (if (eq! 6 (sum-row-1)) 0
        (if (and! (eq! 3 (sum-row-2)) (eq! 1 (multiply-row-2))) 0
            (if (eq! 6 (sum-row-2)) 0
        (if (and! (eq! 3 (sum-row-3)) (eq! 1 (multiply-row-3))) 0
            (if (eq! 6 (sum-row-3)) 0
        (if (and! (eq! 3 (sum-diagonal-left-right)) (eq! 1 (multiply-diagonal-left-right))) 0
            (if (eq! 6 (sum-diagonal-left-right)) 0
        (if (and! (eq! 3 (sum-diagonal-right-left)) (eq! 1 (multiply-diagonal-right-left))) 0
            (if (eq! 6 (sum-diagonal-right-left)) 0
        (if (eq! 13 (sum-boxes)) 0
        (if (eq! 14 (sum-boxes)) 0
        (if (vanishes! (sum-boxes)) 0 1)))))))))))))))))))
)


(defun (check-column-1)
        (if (eq! 3 (sum-column-1))
        (if (eq! 1 (multiply-column-1)) 0 1)
        (if (eq! 6 (sum-column-1)) 0 1))
)

(defun (check-column-2)
        (if (eq! 3 (sum-column-2))
        (if (eq! 1 (multiply-column-2)) 0 1)
        (if (eq! 6 (sum-column-2)) 0 1))
)

(defun (check-column-3)
        (if (eq! 3 (sum-column-3))
        (if (eq! 1 (multiply-column-3)) 0 1)
        (if (eq! 6 (sum-column-3)) 0 1))
)
(defun (check-row-1)
        (if (eq! 3 (sum-row-1))
        (if (eq! 1 (multiply-row-1)) 0 1)
        (if (eq! 6 (sum-row-1)) 0 1))
)
(defun (check-row-2)
        (if (eq! 3 (sum-row-2))
        (if (eq! 1 (multiply-row-2)) 0 1)
        (if (eq! 6 (sum-row-2)) 0 1))
)
(defun (check-row-3)
        (if (eq! 3 (sum-row-3))
        (if (eq! 1 (multiply-row-3)) 0 1)
        (if (eq! 6 (sum-row-3)) 0 1))
)
(defun (check-diagonal-left-right)
        (if (eq! 3 (sum-diagonal-left-right))
        (if (eq! 1 (multiply-diagonal-left-right)) 0 1)
        (if (eq! 6 (sum-diagonal-left-right)) 0 1))
)
(defun (check-diagonal-right-left)
        (if (eq! 3 (sum-diagonal-right-left))
        (if (eq! 1 (multiply-diagonal-right-left)) 0 1)
        (if (eq! 6 (sum-diagonal-right-left)) 0 1))
)

(defun (no-one-has-won-yet-or-finished-or-draw)
 (* (* (* (* (* (* (* (* (* (* (check-column-1) (check-column-2))
                    (check-column-3))
                    (check-row-1))
                    (check-row-2))
                    (check-row-3))
                    (check-diagonal-left-right))
                    (check-diagonal-right-left))
                    (- 13 (sum-boxes)))
                    (- 14 (sum-boxes)))
                    (sum-boxes))
)
;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; Constraints
;;;;;;;;;;;;;;;;;;;;;;;;;;

;; a player cannot play twice in a row
;; the diff-all cannot remain constant (1 -> 1) or (2 ->2), it has to alternate
(defconstraint players-take-turns
                    (:guard (no-one-has-won-yet-or-finished-or-draw))
                    (begin
                        (if (eq! (diff-all) 1) (eq! (diff-all-next) 2))
                        (if (eq! (diff-all) 2) (eq! (diff-all-next) 1))
                    )
)

;; a player cannot change the value of a box is it's already played
;; a non-zero box cannot be changed
(defconstraint player-plays-in-empty-box
                 ()
                 (begin
                    (if-not-zero (prev [BOXES 1]) (remained-constant! [BOXES 1]))
                    (if-not-zero (prev [BOXES 2]) (remained-constant! [BOXES 2]))
                    (if-not-zero (prev [BOXES 3]) (remained-constant! [BOXES 3]))
                    (if-not-zero (prev [BOXES 4]) (remained-constant! [BOXES 4]))
                    (if-not-zero (prev [BOXES 5]) (remained-constant! [BOXES 5]))
                    (if-not-zero (prev [BOXES 6]) (remained-constant! [BOXES 6]))
                    (if-not-zero (prev [BOXES 7]) (remained-constant! [BOXES 7]))
                    (if-not-zero (prev [BOXES 8]) (remained-constant! [BOXES 8]))
                    (if-not-zero (prev [BOXES 9]) (remained-constant! [BOXES 9]))
                 )
)

;; game stops when a player has won
(defconstraint game-is-finished
                 ()
                 (if-not-zero STAMP
                    (if-zero (no-one-has-won-yet-or-finished-or-draw)
                            (vanishes! (sum-boxes-next)))
                 )
)

;; last row has to be either a win, a draw or an empty row
;; should fail if game is mid-way
(defconstraint last-row
                 (:domain {-1})
                 (vanishes! (no-one-has-won-yet-or-finished-or-draw))
)

;; after the game is finished, no one can play
;; once the trace has been 0 once, it stays 0

;; constraint on stamp - not go back to 0

;; this constraint enforces that BOXES in the range [0, 1, 2]
;; TODO: not working
;;(definrange  [BOXES 1]   2)

(defconstraint range-check
                 ()
                 (begin
                    (if-not-eq [BOXES 1] 0 (if-not-eq [BOXES 1] 1 (eq! [BOXES 1] 2)))
                    (if-not-eq [BOXES 2] 0 (if-not-eq [BOXES 2] 1 (eq! [BOXES 2] 2)))
                    (if-not-eq [BOXES 3] 0 (if-not-eq [BOXES 3] 1 (eq! [BOXES 3] 2)))
                    (if-not-eq [BOXES 4] 0 (if-not-eq [BOXES 4] 1 (eq! [BOXES 4] 2)))
                    (if-not-eq [BOXES 5] 0 (if-not-eq [BOXES 5] 1 (eq! [BOXES 5] 2)))
                    (if-not-eq [BOXES 6] 0 (if-not-eq [BOXES 6] 1 (eq! [BOXES 6] 2)))
                    (if-not-eq [BOXES 7] 0 (if-not-eq [BOXES 7] 1 (eq! [BOXES 7] 2)))
                    (if-not-eq [BOXES 8] 0 (if-not-eq [BOXES 8] 1 (eq! [BOXES 8] 2)))
                    (if-not-eq [BOXES 9] 0 (if-not-eq [BOXES 9] 1 (eq! [BOXES 9] 2)))
                 )
)

;; same winning row twice -- check



